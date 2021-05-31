import struct
import numpy as np
import tensorflow as tf
from random import shuffle
from PIL import Image
import matplotlib
import matplotlib.cm as cm
from scipy.special import softmax

nb_samples_per_strain = 2000
nb_strains = 4
nb_samples = nb_strains*nb_samples_per_strain
hash_size = 16

def load_samples(hash_size):
    samples = []
    with open('data/X_CGR_DCT', "rb") as f:
        data = f.read()
        nbfloat = len(data)>>3
        nbSamples = len(data)//(hash_size*hash_size*8)
        for i in range(nbSamples):
            tmp = [0.0 for i in range(hash_size*hash_size)]
            idx = i*8*(hash_size*hash_size)
            buff = data[i*hash_size*hash_size<<3:(i+1)*(hash_size*hash_size)<<3]
            for j in range(hash_size*hash_size):
                    tmp[j] = struct.unpack('d', buff[j*8:(j+1)*8])[0]
            samples += [tmp]
    return np.array(samples)

def train_model():
    #TODO : 
    #Determine how resilient it is to shifting or holes
    #Look at what samples are not correctly classified (about 40 out of 8000), if there is a trend (like large gaps of 'N'),
    #Try without the Hilbert curve to see if it is actually needed
    #Remove irrelevant features to reduce the size of the model
    
    #Load the training data
    X = load_samples(hash_size)

    features = len(X[0])

    #Generates the labels
    Y = []
    Y += [[1, 0, 0, 0] for i in range(nb_samples_per_strain)]
    Y += [[0, 1, 0, 0] for i in range(nb_samples_per_strain)]
    Y += [[0, 0, 1, 0] for i in range(nb_samples_per_strain)]
    Y += [[0, 0, 0, 1] for i in range(nb_samples_per_strain)]

    #Shuffles
    available = [i for i in range(nb_samples)]
    shuffle(available)

    x_train = []
    y_train = []    
    for i in available:
        x_train += [X[i]]
        y_train += [Y[i]]

    x_train = np.array(x_train)
    y_train = np.array(y_train)

    model = tf.keras.Sequential()
    model.add(tf.keras.layers.Dense(4, activation='softmax', kernel_initializer='he_normal', input_shape=(features,)))
    model.compile(
        optimizer='adam',
        loss='binary_crossentropy',
        metrics='binary_accuracy')

    history = model.fit(x_train, y_train, epochs=1000, batch_size=32, verbose=2, validation_split=0.0)

    predictions = model.predict(x_train, verbose=2)

    err = 0
    for j in range(len(x_train)):
        L = predictions[j]
        Y = [i for i in L].index(max([i for i in L]))
        if Y != [i for i in y_train[j]].index(max(y_train[j])):
            err += 1
    acc = 1-(err/8000)
    print(acc)

    for i in range(len(model.layers)):
        layer = model.layers[i]
        #print(layer.get_config())
        weights = layer.get_weights()

        norm = matplotlib.colors.Normalize(vmin=np.min(weights[0]), vmax=np.max(weights[0]), clip=True)
        mapper = cm.ScalarMappable(norm=norm, cmap=cm.seismic)

        array = np.array([[[round(i*255) for i in mapper.to_rgba(i)] for i in i] for i in weights[0]], dtype=np.uint8)
        new_image = Image.fromarray(array)
        new_image.save('model/weights_layer_{}.png'.format(i))

        #Save both bias and weights in bianry and npy format
        data = []
        flattened_weights = weights[0].flatten()
        for j in range(len(flattened_weights)):
            data += struct.pack('d', flattened_weights[j])

        with open("model/weights_layer_{}".format(i), "wb") as f:
            f.write(bytearray(data))
            f.close()
        np.save('model/weights_layer_{}.npy'.format(i), weights[0])
        
        data = []
        flattened_bias = weights[1].flatten()
        for j in range(len(flattened_bias)):
            data += struct.pack('d', flattened_bias[j])

        with open("model/bias_layer_{}".format(i), "wb") as f:
            f.write(bytearray(data))
            f.close()
        np.save('model/bias_layer_{}.npy'.format(i), weights[1])

train_model()
